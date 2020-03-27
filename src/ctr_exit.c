#define _GNU_SOURCE

#include "ctr_exit.h"
#include "cli.h" // opt_exit_command
#include "utils.h"
#include "parent_pipe_fd.h"
#include "globals.h"

#include <errno.h>
#include <glib.h>
#include <glib-unix.h>
#include <signal.h>
#include <stdlib.h>
#include <unistd.h>

volatile pid_t container_pid = -1;
volatile pid_t create_pid = -1;

void on_sigchld(G_GNUC_UNUSED int signal)
{
	ntracef("start function: %s", __FUNCTION__);
	raise(SIGUSR1);
}

void on_sig_exit(int signal)
{
	ntracef("start function: %s", __FUNCTION__);
	if (container_pid > 0) {
		ntracef("function: %s container_pid %d; killing", __FUNCTION__, container_pid);
		if (kill(container_pid, signal) == 0)
			return;
	} else if (create_pid > 0) {
		ntracef("function: %s create_pid %d; killing", __FUNCTION__, create_pid);
		if (kill(create_pid, signal) == 0)
			return;
		ntracef("function: %s failed", __FUNCTION__);
		if (errno == ESRCH) {
			ntracef("function: %s esrch", __FUNCTION__);
			/* The create_pid process might have exited, so try container_pid again.  */
			if (container_pid > 0 && kill(container_pid, signal) == 0)
				return;
		}
	}
	/* Just force a check if we get here.  */
	raise(SIGUSR1);
}

void check_child_processes(GHashTable *pid_to_handler)
{
	ntracef("start function: %s", __FUNCTION__);
	for (;;) {
		int status;
		pid_t pid = waitpid(-1, &status, WNOHANG);

		if (pid < 0 && errno == EINTR) {
			ntracef("eintr: %s", __FUNCTION__);
			continue;
		}

		if (pid < 0 && errno == ECHILD) {
			ntracef("quitting from echild %s", __FUNCTION__);
			g_main_loop_quit(main_loop);
			return;
		}
		if (pid < 0)
			pexit("Failed to read child process status");

		if (pid == 0)
			return;

		/* If we got here, pid > 0, so we have a valid pid to check.  */
		void (*cb)(GPid, int, gpointer) = g_hash_table_lookup(pid_to_handler, &pid);
		if (cb) {
			ntracef("calling cb %s", __FUNCTION__);
			cb(pid, status, 0);
		} else if (opt_api_version >= 1) {
			ndebugf("couldn't find cb for pid %d", pid);
			if (container_status < 0 && container_pid < 0 && opt_exec && opt_terminal) {
				ndebugf("container status and pid were found prior to callback being registered. calling manually");
				container_exit_cb(pid, status, 0);
			}
		}
	}
}

gboolean on_sigusr1_cb(gpointer user_data)
{
	ntracef("start function: %s", __FUNCTION__);
	GHashTable *pid_to_handler = (GHashTable *)user_data;
	check_child_processes(pid_to_handler);
	return G_SOURCE_CONTINUE;
}

gboolean timeout_cb(G_GNUC_UNUSED gpointer user_data)
{
	ntracef("start function: %s", __FUNCTION__);
	timed_out = TRUE;
	ninfo("Timed out, killing main loop");
	g_main_loop_quit(main_loop);
	return G_SOURCE_REMOVE;
}

int get_exit_status(int status)
{
	ntracef("start function: %s", __FUNCTION__);
	if (WIFEXITED(status))
		return WEXITSTATUS(status);
	if (WIFSIGNALED(status))
		return 128 + WTERMSIG(status);
	return -1;
}

void runtime_exit_cb(G_GNUC_UNUSED GPid pid, int status, G_GNUC_UNUSED gpointer user_data)
{
	ntracef("start function: %s", __FUNCTION__);
	runtime_status = status;
	create_pid = -1;
	g_main_loop_quit(main_loop);
}

void container_exit_cb(G_GNUC_UNUSED GPid pid, int status, G_GNUC_UNUSED gpointer user_data)
{
	ntracef("start function: %s", __FUNCTION__);
	if (get_exit_status(status) != 0) {
		ninfof("container %d exited with status %d", pid, get_exit_status(status));
	}
	container_status = status;
	container_pid = -1;
	/* In the case of a quickly exiting exec command, the container exit callback
	   sometimes gets called earlier than the pid exit callback. If we quit the loop at that point
	   we risk falsely telling the caller of conmon the runtime call failed (because runtime status
	   wouldn't be set). Instead, don't quit the loop until runtime exit is also called, which should
	   shortly after. */
	if (opt_api_version >= 1 && create_pid > 0 && opt_exec && opt_terminal) {
		ndebugf("container pid return handled before runtime pid return. Not quitting yet.");
		return;
	}

	g_main_loop_quit(main_loop);
}

void do_exit_command()
{
	ntracef("start function: %s", __FUNCTION__);
	if (sync_pipe_fd > 0) {
		close(sync_pipe_fd);
		sync_pipe_fd = -1;
	}

	if (signal(SIGCHLD, SIG_DFL) == SIG_ERR) {
		_pexit("Failed to reset signal for SIGCHLD");
	}

	pid_t exit_pid = fork();
	if (exit_pid < 0) {
		_pexit("Failed to fork");
	}

	if (exit_pid) {
		int ret, exit_status = 0;

		do
			ret = waitpid(exit_pid, &exit_status, 0);
		while (ret < 0 && errno == EINTR);

		if (exit_status)
			_exit(exit_status);

		return;
	}

	/* Count the additional args, if any.  */
	size_t n_args = 0;
	if (opt_exit_args)
		for (; opt_exit_args[n_args]; n_args++)
			;

	gchar **args = malloc(sizeof(gchar *) * (n_args + 2));
	if (args == NULL)
		_exit(EXIT_FAILURE);

	args[0] = opt_exit_command;
	if (opt_exit_args)
		for (n_args = 0; opt_exit_args[n_args]; n_args++)
			args[n_args + 1] = opt_exit_args[n_args];
	args[n_args + 1] = NULL;

	execv(opt_exit_command, args);

	/* Should not happen, but better be safe. */
	_exit(EXIT_FAILURE);
}
