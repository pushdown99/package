#include <Windows.h>
#include <iostream>
#include <stdbool.h>
#include <stdlib.h>
#include <WinBase.h>
using namespace std;

typedef int (__stdcall *MainAFn)(const char *pProgName, char *pCmdLine);

int CreateVSPair(int nPair, char *Port1, char *Port2) {
    OSVERSIONINFO VersionInfo;
    HINSTANCE libInst;
    BOOL fFreeResult, fRunTimeLinkSuccess = FALSE; 
    int returnvalue = false;

    VersionInfo.dwOSVersionInfoSize = sizeof(OSVERSIONINFO);
    GetVersionEx(&VersionInfo);

    if (VersionInfo.dwPlatformId != VER_PLATFORM_WIN32_NT) {
        printf("setup.dll %d %d\n", VersionInfo.dwPlatformId, VER_PLATFORM_WIN32_NT);
        libInst = LoadLibrary("setup.dll");
    }
    else {
        printf("setup.dll %d %d\n", VersionInfo.dwPlatformId, VER_PLATFORM_WIN32_NT);
        libInst = LoadLibrary(TEXT("setup.dll"));
    }

    MainAFn mainA=(MainAFn)GetProcAddress(libInst, "MainA");
    if (mainA == 0) {
        printf("Couldn't find function\n");
        return false; /* Couldn't find function */
    }
    char pCmdLine[128];
    printf("Port %s %s\n", Port1, Port2);
    sprintf(pCmdLine, "install %d PortName=%s PortName=%s", nPair, Port1, Port2);
    returnvalue = mainA("com0com", pCmdLine); 
    FreeLibrary(libInst);
    return returnvalue;
}

int DeleteVSPair(int nPair) {
    OSVERSIONINFO VersionInfo;
    HINSTANCE libInst;
    BOOL fFreeResult, fRunTimeLinkSuccess = FALSE; 
    int returnvalue = false;

    VersionInfo.dwOSVersionInfoSize = sizeof(OSVERSIONINFO);
    GetVersionEx(&VersionInfo);

    if (VersionInfo.dwPlatformId != VER_PLATFORM_WIN32_NT) {
        printf("setup.dll %d %d\n", VersionInfo.dwPlatformId, VER_PLATFORM_WIN32_NT);
        libInst = LoadLibrary("setup.dll");
    }
    else {
        printf("setup.dll %d %d\n", VersionInfo.dwPlatformId, VER_PLATFORM_WIN32_NT);
        libInst = LoadLibrary(TEXT("setup.dll"));
    }

    MainAFn mainA=(MainAFn)GetProcAddress(libInst, "MainA");
    if (mainA == 0) {
        printf("Couldn't find function\n");
        return false; /* Couldn't find function */
    }
    char pCmdLine[128];
    sprintf(pCmdLine, "remove %d", nPair);
    returnvalue = mainA("com0com", pCmdLine); 
    FreeLibrary(libInst);
    return returnvalue;
}

int main(int argc,char* argv[])
{
    if(argc < 2) {
        printf ("\n");
        printf ("Usage: %s <coammand>\n\n", argv[0]);
        printf ("Commands:\n");
        printf ("  install <n> <prmsA> <prmsB>\n");
        printf ("  remove <n>");
        printf ("\n");
        exit(0);
    }
    if (!strcmp(argv[1], "install")) {
        CreateVSPair (atoi(argv[2]), argv[3], argv[4]);
    }
    else {
        DeleteVSPair (atoi(argv[2]));
    }
}
