#include "utils.h"
#include <string.h>
#include <strings.h>

log_level_t log_level = WARN_LEVEL;
char *log_cid = NULL;
gboolean use_syslog = FALSE;
FILE *output_file = NULL;

/* Set the log level for this call. log level defaults to warning.
   parse the string value of level_name to the appropriate log_level_t enum value
*/
void set_output_file(FILE *fp) {
	output_file = fp;
}
void set_conmon_logs(char *level_name, char *cid_, gboolean syslog_, char *tag, char *opt_output_file_, FILE **conmon_output_file_)
{
	if (tag == NULL)
		log_cid = cid_;
	else
		log_cid = g_strdup_printf("%s: %s", cid_, tag);

	output_file = stderr;
	if (opt_output_file_ != NULL) {
		output_file = fopen(opt_output_file_, "a");
		*conmon_output_file_ = output_file;
	}


	use_syslog = syslog_;
	// log_level is initialized as Warning, no need to set anything
	if (level_name == NULL)
		return;
	if (!strcasecmp(level_name, "error") || !strcasecmp(level_name, "fatal") || !strcasecmp(level_name, "panic")) {
		log_level = EXIT_LEVEL;
		return;
	} else if (!strcasecmp(level_name, "warn") || !strcasecmp(level_name, "warning")) {
		log_level = WARN_LEVEL;
		return;
	} else if (!strcasecmp(level_name, "info")) {
		log_level = INFO_LEVEL;
		return;
	} else if (!strcasecmp(level_name, "debug")) {
		log_level = DEBUG_LEVEL;
		return;
	} else if (!strcasecmp(level_name, "trace")) {
		log_level = TRACE_LEVEL;
		return;
	}
	ntracef("set log level to %s", level_name);
	nexitf("No such log level %s", level_name);
}

ssize_t write_all(int fd, const void *buf, size_t count)
{
	size_t remaining = count;
	const char *p = buf;
	ssize_t res;

	while (remaining > 0) {
		do {
			res = write(fd, p, remaining);
		} while (res == -1 && errno == EINTR);

		if (res <= 0)
			return -1;

		remaining -= res;
		p += res;
	}

	return count;
}
