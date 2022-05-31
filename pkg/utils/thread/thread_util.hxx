#ifndef __THREAD_UTIL_H__
#define __THREAD_UTIL_H__

#include <string>
#include <unordered_map>
#include <thread>
#include <mutex>

using namespace std;

class ThreadUtil
{
  public:
    static string* get_thread_id();

  private:
    static long thread_counter;
    static bool inited;
    static unordered_map<string, string*> tids;
    static mutex thread_mutex;

    // we control our own init
    static void init();
    static string* create_thread_name();

    // not exposed as I don't know how to manage the memory
    // outside the c/c++ world
    static string get_log_tm();
};

#endif
