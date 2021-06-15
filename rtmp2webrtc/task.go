// Copyright 2021, Chef.  All rights reserved.
// https://github.com/q191201771/lalext
//
// Use of this source code is governed by a MIT-style license
// that can be found in the License file.
//
// Author: Chef (191201771@qq.com)

package main

import (
	"github.com/pion/webrtc/v3"
	"github.com/q191201771/lal/pkg/avc"
	"github.com/q191201771/lal/pkg/base"
	"github.com/q191201771/lal/pkg/rtmp"
	"github.com/q191201771/naza/pkg/nazalog"
)

func StartTunnelTask(rtmpUrl string, sessionDescription string, onLocalSessDesc func(sessDesc string)) error {
	var (
		iceConnectionStateChan = make(chan webrtc.ICEConnectionState, 1)
	)

	var sd webrtc.SessionDescription
	err := decodeBase64Json(sessionDescription, &sd)
	if err != nil {
		return err
	}

	var webrtcSender WebRtcSender
	localSessionDescription, err := webrtcSender.Init(sd, func(connectionState webrtc.ICEConnectionState) {
		nazalog.Debugf("> OnICEConnectionStateChange. state=%s", connectionState.String())
		switch connectionState {
		case webrtc.ICEConnectionStateChecking:
			// noop
		case webrtc.ICEConnectionStateConnected:
			iceConnectionStateChan <- connectionState
		case webrtc.ICEConnectionStateFailed:
			iceConnectionStateChan <- connectionState
		default:
			nazalog.Errorf("NOTICEME %s", connectionState.String())
		}
	})
	if err != nil {
		return err
	}
	encLocalSessionDescription, err := encodeBase64Json(localSessionDescription)
	if err != nil {
		return err
	}
	onLocalSessDesc(encLocalSessionDescription)

	ics := <-iceConnectionStateChan
	if ics == webrtc.ICEConnectionStateFailed {
		return ErrRtmp2WebRtc
	}

	pullSession := rtmp.NewPullSession(func(option *rtmp.PullSessionOption) {
	})

	prevTimestamp := uint32(0)
	var sps []byte
	var pps []byte
	err = pullSession.Pull(rtmpUrl, func(msg base.RtmpMsg) {
		switch msg.Header.MsgTypeId {
		case base.RtmpTypeIdMetadata:
			// noop
			return
		case base.RtmpTypeIdAudio:
			// audio not support yet
			return
		case base.RtmpTypeIdVideo:
			codecid := msg.Payload[0] & 0xF
			if codecid != base.RtmpCodecIdAvc {
				nazalog.Errorf("invalid codec id since only support avc yet. codecid=%d", codecid)
				return
			}
		}

		//nazalog.Tracef("< R len=%d, type=%s, hex=%s", len(msg.Payload), avc.ParseNaluTypeReadable(msg.Payload[5+4]), hex.Dump(nazastring.SubSliceSafety(msg.Payload, 16)))

		timestamp := msg.Header.TimestampAbs

		// 只缓存
		if msg.IsAvcKeySeqHeader() {
			sps, pps, err = avc.ParseSpsPpsFromSeqHeader(msg.Payload)
			nazalog.Assert(nil, err)
			nazalog.Debugf("cache spspps. sps=%d, pps=%d", len(sps), len(pps))
			return
		}

		if prevTimestamp == 0 {
			prevTimestamp = timestamp
		}

		var out []byte
		err = avc.IterateNaluAvcc(msg.Payload[5:], func(nal []byte) {
			t := avc.ParseNaluType(nal[0])
			//nazalog.Debugf("iterate nalu. type=%s, len=%d", avc.ParseNaluTypeReadable(nal[0]), len(nal))
			if t == avc.NaluTypeSei {
				return
			}

			if t == avc.NaluTypeIdrSlice {
				out = append(out, avc.NaluStartCode3...)
				out = append(out, sps...)
				out = append(out, avc.NaluStartCode3...)
				out = append(out, pps...)
				nazalog.Debugf("append spspps. sps=%d, pps=%d", len(sps), len(pps))
			}

			out = append(out, avc.NaluStartCode3...)
			out = append(out, nal...)
		})

		if len(out) != 0 {
			_ = webrtcSender.Write(out, timestamp-prevTimestamp)
		}

		prevTimestamp = timestamp
	})
	if err != nil {
		return err
	}
	return <-pullSession.WaitChan()
}