#include "Precomp.h"
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include "CpuArch.h"
#include "Alloc.h"
#include "7zFile.h"
#include "7zVersion.h"
#include "LzmaDec.h"
#include "LzmaEnc.h"

#define IN_BUF_SIZE (1 << 16)
#define OUT_BUF_SIZE (1 << 16)

int lzmaCompressUtil(const char *src,const char *dst,int encodeMode);
