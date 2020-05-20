#include <stddef.h>
#include <fcntl.h>
#include <getopt.h>
#include <linux/limits.h>
#include <signal.h>
#include <stdbool.h>
#include <stdio.h>
#include <stdlib.h>
#include <sys/mount.h>
#include <sys/stat.h>
#include <sys/types.h>
#include <unistd.h>
#include "pinns.h"
#include "utils.h"

static int bind_ns(const char *pin_path, const char *filename, const char *ns_name);
static int directory_exists_or_create(const char* path);

int bind_pid_ns(const char *pin_path, const char *filename) {
    if (bind_ns(pin_path, filename, "pid") < 0) {
      return -1;
    }
	return 0;
}

static int bind_ns(const char *pin_path, const char *filename, const char *ns_name) {
  char bind_path[PATH_MAX];
  char ns_path[PATH_MAX];
  int fd;

  // first, verify the /$PATH/$NSns directory exists
  snprintf(bind_path, PATH_MAX - 1, "%s/%sns", pin_path, ns_name);
  if (directory_exists_or_create(bind_path) < 0) {
    nwarnf("%s exists and is not a directory", bind_path);
    return -1;
  }

  // now, get the real path we want
  snprintf(bind_path, PATH_MAX - 1, "%s/%sns/%s", pin_path, ns_name, filename);

  fd = open(bind_path, O_RDONLY | O_CREAT | O_EXCL, 0);
  if (fd < 0) {
    pwarn("Failed to create ns file");
    return -1;
  }
  close(fd);

  snprintf(ns_path, PATH_MAX - 1, "/proc/self/ns/%s", ns_name);
  if (mount(ns_path, bind_path, NULL, MS_BIND, NULL) < 0) {
    nwarnf("Failed to bind mount ns: %s", ns_path);
    return -1;
  }

  return 0;
}

static int directory_exists_or_create(const char* path) {
  struct stat sb;
  if (stat(path, &sb) != 0) {
    mkdir(path, 0755);
	return 0;
  }

  if (!S_ISDIR(sb.st_mode)) {
    return -1;
  }
  return 0;
}
