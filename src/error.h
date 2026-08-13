#ifndef _ERROR_H
#define _ERROR_H

#include <setjmp.h>

// jump target for fatalf; set this up with setjmp() before running any code
// that might call fatalf, then check its return value to detect a fatal error.
extern jmp_buf fatal_jmp_buf;

#ifndef ERRORF_ENABLED
#define ERRORF_ENABLED 1
#endif
#ifndef INFOF_ENABLED
#define INFOF_ENABLED 1
#endif

#if ERRORF_ENABLED
int errorf(const char *format, ...);
#else
#define errorf(...) ((void)0)
#endif

#if INFOF_ENABLED
int infof(const char *format, ...);
#else
#define infof(...) ((void)0)
#endif

int fatalf(const char *format, ...);

#endif