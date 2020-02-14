#include <string.h>
#include <stdio.h>
#include <stdlib.h>
#include <unistd.h>

#define CR 13

void spoof_oom() {
	char *p;  

	sleep(10);
	while (1) {  
		if ((p = malloc(1<<20)) == NULL) {  
			return;  
		}
		memset (p, 0, (1<<20));  
		sleep(.0001);
	}  
}
