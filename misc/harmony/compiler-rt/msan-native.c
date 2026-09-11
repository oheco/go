#include <stdlib.h>
#include <stdio.h>
int main(int argc, char **argv) {
  int *p = argc > 1 ? malloc(sizeof(int)) : calloc(1, sizeof(int));
  if (!p) return 2;
  printf("%d\n", *p);
  free(p);
  return 0;
}
