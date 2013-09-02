#!/bin/sh

llvm-as | opt /dev/stdin $* | llvm-dis
