// Copyright 2021, Chef.  All rights reserved.
// https://github.com/q191201771/lalext
//
// Use of this source code is governed by a MIT-style license
// that can be found in the License file.
//
// Author: Chef (191201771@qq.com)
#ifndef _LALCC_FORWARD_DECLARATION_HPP_
#define _LALCC_FORWARD_DECLARATION_HPP_

// lalcc_forward_decalaration.hpp
//
// 前向声明，供liblalcc内部各文件包含
//

#include <string>
#include <sstream>

extern "C" {
#include "libavformat/avformat.h"
#include "libswresample/swresample.h"
#include "libavutil/avutil.h"
#include "libavutil/timestamp.h"
#include "libavutil/opt.h"
}

#include "chef_base/chef_env.hpp"
#include "chef_base/chef_task_thread.hpp"
#include "chef_base/chef_wait_event_counter.hpp"
#include "chef_base/chef_stuff_op.hpp"

namespace lalcc {

class OpenTimeoutHooker;

class AvPacketT;
typedef chef::shared_ptr<AvPacketT> AvPacketTPtr;

class PullSession;
typedef chef::shared_ptr<PullSession> PullSessionPtr;

class PushSession;
typedef chef::shared_ptr<PushSession> PushSessionPtr;

class Decode;
typedef chef::shared_ptr<Decode> DecodePtr;
}

#endif