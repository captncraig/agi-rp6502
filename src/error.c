#include "error.h"
#include <stdarg.h>
#include <stdio.h>

#if ERRORF_ENABLED
int errorf(const char *format, ...){
    printf("\x1b[32mERROR:\x1b[0m ");
    va_list args;
    va_start(args, format);
    int n = vprintf(format, args);
    va_end(args);
    return n;
}
#endif

#if INFOF_ENABLED
int infof(const char *format, ...){
    printf("\x1b[34mINFO:\x1b[0m ");
    va_list args;
    va_start(args, format);
    int n = vprintf(format, args);
    va_end(args);
    return n;
}
#endif

int fatalf(const char *format, ...){
    printf("\x1b[33mFATAL:\x1b[0m ");
    va_list args;
    va_start(args, format);
    int n = vprintf(format, args);
    va_end(args);
    return n;
}