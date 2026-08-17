#ifndef _ERROR_H
#define _ERROR_H

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