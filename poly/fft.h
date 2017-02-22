#ifndef _FFT__H
#define _FFT__H

#include <stdbool.h>
#include <flint/fmpz_mod_poly.h>

char *fft_interpolate(char *modIn, int nPoints, 
    char **rootsIn, char **pointsYin, bool invert);

#endif
