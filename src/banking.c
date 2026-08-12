#include "banking.h"
#include <rp6502.h>
#include <stdio.h>

void init_banks(){
    VIA.ddrb = 0x3f; //6 bits of port b are bank number
    VIA.prb = 0x00;
}
unsigned char change_bank(unsigned char bankNum){
    unsigned char oldPrb = VIA.prb;
    VIA.prb = (oldPrb & 0xc0) | (bankNum & 0x3f); 
    return oldPrb & 0x3f;
}

#define BANK_STACK_SIZE 8
unsigned char bankStack[BANK_STACK_SIZE];
unsigned char bankStackPointer = 0;

void push_bank(unsigned char bankNum){
    unsigned char oldBank = change_bank(bankNum);
    if (bankStackPointer < BANK_STACK_SIZE) {
        bankStack[bankStackPointer++] = oldBank;
    }else{
        printf("Bank stack overflow!!!!!!\n");
    }
}
void pop_bank(){
    if (bankStackPointer > 0) {
        bankStackPointer--;
        change_bank(bankStack[bankStackPointer]);
    }else{
         printf("Bank stack underflow!!!!!!\n");
    }
}