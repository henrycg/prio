#include <stdio.h>
#include <stdlib.h>
#include <flint/fmpz.h>

#define DEBUG_INT(t,a) do {\
printf("%s: ", t);\
fmpz_print(a);\
printf("\n");\
} while(0);

inline void *
safe_malloc(size_t bytes)
{
  void *out = malloc(bytes);
  if (!out) {
    fprintf(stderr, "Malloc failed\n");
    exit(1);
  }
  return out;
}

static inline void
fmpz_init_from_gostr(fmpz_t out, char *in)
{
  fmpz_init(out);
  fmpz_set_str(out, in, 16);
  free(in);
}

char *fmpz_array_to_str(int n, fmpz_t *input); 

