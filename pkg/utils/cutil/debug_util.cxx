#include "debug_util.hxx"
#include "thread_util.hxx"

//global var, not thread safe!
string g_result;

const char* getThreadId() {
  g_result = ThreadUtil::get_thread_id();
  return g_result.c_str();
}

const char* getLogTm() {
  g_result = ThreadUtil::get_log_tm().c_str();
  return g_result.c_str();
}
