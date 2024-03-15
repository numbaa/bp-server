# bp-server

A naive [breakpad](https://chromium.googlesource.com/breakpad/breakpad) server.

## Usage
1. Download from Github Releases or build it yourself.
```bash
$> ./bp-server -c /path/to/bp-server.xml
$> # or just
$> ./bp-server
```

2. Put [minidump_stackwalk](https://github.com/numbaa/breakpad-build/releases) in `$PATH`.

3. Make symbol file from your exe/pdb using [dump_syms](https://github.com/mozilla/dump_syms).
```bash
$> dump_syms --store ./symbols /path/to/your-app.pdb
```

4. Upload your symbol file
```bash
$> curl -X POST \
    -H "Content-Type: multipart/form-data" \
    -F "file=@./symbols/your-app.pdb/123123123123123/your-app.sym" \
    -F "entry=your-app.pdb" \
    -F "id=123123123123123" \
    http://your-host:17001/upsym
```

5. Send dmp file when your program crash
```c++
static bool minidump_callback(
    const wchar_t* dump_path,
    const wchar_t* minidump_id,
    void* context,
    EXCEPTION_POINTERS* exinfo,
    MDRawAssertionInfo* assertion,
    bool succeeded)
{
    std::map<std::wstring, std::wstring> parameters;
    std::map<std::wstring, std::wstring> files;
    int timeout = 1000;
    int response_code = 0;
    std::wstring response_body;
    parameters[L"build"] = L"" __DATE__ " " __TIME__;
    parameters[L"os"] = L"Windows";
    parameters[L"program"] = L"your-app.exe";
    parameters[L"version"] = L"v3.2.1";
    std::wstring fullpath;
    fullpath = fullpath + dump_path + L"/" + minidump_id + L".dmp";
    files[L"file"] = fullpath;
    google_breakpad::HTTPUpload::SendMultipartPostRequest(L"http://your-host:17001/updump", parameters, files, &timeout, &response_body, &response_code);
    return false;
}
```

5. Visit `http://your-host:17000/list/{page}`
```
http://your-host:17000/list/0
```
And you get
|  ID |  OS | Program     |   Version   |  Build Time |  Crash Time |    Dump     |
| --- | --- | ----------- | ----------- | ----------- | ----------- | ----------- |
| 2 |Windows| your-app.exe|   v3.2.1    |Mar 13 2024 02:16:47| 2024-03-13 03:47:46.8633922| [12345678-1234-5678-9876-123456789abc.dmp]() |
| 1 |Windows| your-app.exe|   v3.2.1    |Mar 13 2024 02:16:47| 2024-03-13 10:31:12.6846541| [87654321-4321-8765-6789-cba987654321.dmp]() |

6. Click the dump links and you get something like this
```plaintext
Operating system: Windows NT
                  10.0.22621 
CPU: amd64
     family 25 model 33 stepping 0
     12 CPUs

GPU: UNKNOWN

Crash reason:  EXCEPTION_ACCESS_VIOLATION_WRITE
Crash address: 0x0
Process uptime: 6 seconds

Thread 9 (crashed)
 0  lanthing-app.exe!(anonymous namespace)::crash_me() [threads.cpp : 98 + 0x0]
    rax = 0x0000000000000000   rdx = 0x000002946dc80000
    rcx = 0x0000000000000001   rbx = 0x0000029474e1c6e0
    rsi = 0x0000029474f76c40   rdi = 0x000000000000003a
    rbp = 0x000000a294bffb10   rsp = 0x000000a294bffa08
     r8 = 0x000002946db92280    r9 = 0x0000000000000001
    r10 = 0x0000000000000003   r11 = 0x000000a294bff910
    r12 = 0x00000000000002bc   r13 = 0x000002946dcf1020
    r14 = 0x0000000000000003   r15 = 0x0000000002b13666
    rip = 0x00007ff6f51ecc12
    Found by: given as instruction pointer in context
 1  lanthing-app.exe!ltlib::ThreadWatcher::checkLoop() [threads.cpp : 183 + 0x5]
    rbp = 0x000000a294bffb10   rsp = 0x000000a294bffa10
    rip = 0x00007ff6f51ec865
    Found by: stack scanning
 2  lanthing-app.exe!std::thread::_Invoke<std::tuple<std::_Binder<std::_Unforced,void (__cdecl ltlib::ThreadWatcher::*)(void),ltlib::ThreadWatcher *> >,0>(void*) [thread : 60 + 0x6]
    rbx = 0x000002946dcf99c0   rsi = 0x0000000000000000
    rdi = 0x0000000000000000   rbp = 0x0000000000000000
    rsp = 0x000000a294bffbf0   r12 = 0x0000000000000000
    r13 = 0x0000000000000000   r14 = 0x0000000000000000
    r15 = 0x0000000000000000   rip = 0x00007ff6f51e5acf
    Found by: call frame info
 3  ucrtbase.dll + 0x29363
    rbx = 0x000002946dcf2710   rbp = 0x0000000000000000
    rsp = 0x000000a294bffc20   r12 = 0x0000000000000000
    r13 = 0x0000000000000000   r14 = 0x0000000000000000
    r15 = 0x0000000000000000   rip = 0x00007ff8e7769363
    Found by: call frame info
......
```
