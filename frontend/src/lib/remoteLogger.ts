type LogLevel = 'debug' | 'info' | 'warn' | 'error';

class RemoteLogger {
  private async send(level: LogLevel, message: string, context?: any) {
    try {
      // Don't log if we're on the server (e.g. during SSR)
      if (typeof window === 'undefined') return;

      const body = {
        level,
        message,
        context: {
          url: window.location.href,
          userAgent: navigator.userAgent,
          ...context,
        },
      };

      // We use a background fetch so it doesn't block the UI
      fetch('/api/logs', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(body),
        keepalive: true, // Ensure the log is sent even if the page is unloaded
      }).catch((err) => {
        // Silent fail for the network request itself to avoid infinite loops
      });
    } catch (err) {
      // Fallback to console if everything fails
      console.error('RemoteLogger critical failure:', err);
    }
  }

  debug(message: string, context?: any) {
    if (process.env.NODE_ENV === 'development') {
      this.logToConsole('debug', message, context);
    }
    this.send('debug', message, context);
  }

  info(message: string, context?: any) {
    this.logToConsole('info', message, context);
    this.send('info', message, context);
  }

  warn(message: string, context?: any) {
    this.logToConsole('warn', message, context);
    this.send('warn', message, context);
  }

  error(message: string, context?: any) {
    this.logToConsole('error', message, context);
    this.send('error', message, context);
  }

  private logToConsole(level: LogLevel, message: string, context?: any) {
    const prefix = `[RemoteLog:${level}]`;
    const args: any[] = [prefix, message];
    if (context && Object.keys(context).length > 0) {
      args.push(context);
    }

    // We use console.info for local display of remote logs
    // to keep them visible but non-intrusive (no red text or stack-trace-bloat).
    console.info(...args);
  }
}

export const remoteLogger = new RemoteLogger();
