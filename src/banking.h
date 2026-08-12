#ifndef _BANKING_H
#define _BANKING_H

void init_banks();
unsigned char change_bank(unsigned char bankNum);
void push_bank(unsigned char bankNum);
void pop_bank();

#endif