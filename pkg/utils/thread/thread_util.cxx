#include "thread_util.hxx"

#include <chrono>
#include <string>
#include <sstream>
#include <iomanip>
#include <iostream>
#include <thread>

using namespace std;

long ThreadUtil::thread_counter = 0;
bool ThreadUtil::inited = false;
unordered_map<string, string*> ThreadUtil::tids;
mutex ThreadUtil::thread_mutex;

string ThreadUtil::get_log_tm()
{
  ostringstream output;
  time_t currentTime = time(nullptr);
  output << put_time(localtime(&currentTime), "%Y-%m%d:%H:%M:%S-");

  auto now(chrono::system_clock::now());
  auto seconds_since_epoch(
      chrono::duration_cast<chrono::seconds>(now.time_since_epoch()));

  // Construct time_t using 'seconds_since_epoch' rather than 'now' since it is
  // implementation-defined whether the value is rounded or truncated.
  time_t now_t(chrono::system_clock::to_time_t(chrono::system_clock::time_point(seconds_since_epoch)));
  output << (now.time_since_epoch() - seconds_since_epoch).count() / 1000000;
  return output.str();
}

void ThreadUtil::init()
{
  if (!inited)
  {
    return;
  }
  thread_counter = 0;
  tids.clear();
  inited = true;
}

string* ThreadUtil::get_thread_id()
{
  unique_lock<mutex> lock(thread_mutex);
  init();
  ostringstream rec_buf;
  thread::id cur_tid = this_thread::get_id();
  rec_buf << cur_tid;
  string idstr = rec_buf.str();
  // cout << "current calling id: " << idstr << endl;
  auto iter = tids.find(idstr);
  if (iter == tids.end())
  {
    //no found
    string* t_name = create_thread_name();
    tids[idstr] = t_name;
    return t_name;
  }
  //found it
  return iter->second;
}

string* ThreadUtil::create_thread_name()
{
  ostringstream ots;
  ots << "thread-" << thread_counter++;
  string* new_name = new string(ots.str());
  return new_name;
}

string toString(long value)
{
  string number;
  stringstream strstream;
  strstream << value;
  strstream >> number;
  return number;
}
