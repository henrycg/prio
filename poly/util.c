
#include "util.h"

char *fmpz_array_to_str(int n_points, fmpz_t *input) 
{
  // Leave space for final NULL terminator
  size_t outlen = 1;
  for (int i = 0; i < n_points; i++) {
    outlen += fmpz_sizeinbase(input[i], 16) + 2;
  }

  char *out = safe_malloc(outlen * sizeof(char));
  char *outp = out;
  for (int i = 0; i < n_points; i++) {
    fmpz_get_str(outp, 16, input[i]);
    fmpz_clear(input[i]);
  
    // Advance pointer until after \0 terminator  
    do outp++; while(*outp);
    *outp = '\n';
    outp++;
  }
  *outp = '\0';

  return out;
}
