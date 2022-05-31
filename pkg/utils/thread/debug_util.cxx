#include "debug_util.hxx"
#include "thread_util.hxx"

const char* getThreadId() {
  return ThreadUtil::get_thread_id()->c_str();
}
