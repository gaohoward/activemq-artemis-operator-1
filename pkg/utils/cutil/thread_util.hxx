#ifndef __THREAD_UTIL_H__
#define __THREAD_UTIL_H__

#include <string>

using namespace std;

class ThreadUtil
{
public:
  static string get_thread_id();
  static string get_log_tm();
};

#endif
