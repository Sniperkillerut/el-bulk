type LogLevel = 'trace' | 'debug' | 'info' | 'warn' | 'error' | 'off';

const LEVEL_MAP: Record<LogLevel, number> = {
  trace: 0,
  debug: 1,
  info: 2,
  warn: 3,
  error: 4,
  off: 5,
};

class RemoteLogger {
  private currentLevel: LogLevel = 'info';

  constructor() {
    if (typeof window !== 'undefined') {
      const saved = localStorage.getItem('EL_BULK_LOG_LEVEL') as LogLevel;
      if (saved && LEVEL_MAP[saved] !== undefined) {
        this.currentLevel = saved;
      } else {
        this.currentLevel = (process.env.LOG_LEVEL?.toLowerCase() as LogLevel) || 'info';
      }
    }
  }

  setLevel(level: LogLevel) {
    this.currentLevel = level;
    if (typeof window !== 'undefined') {
      localStorage.setItem('EL_BULK_LOG_LEVEL', level);
      console.info(`[RemoteLogger] Level changed to: ${level.toUpperCase()}`);
    }
  }

  private async send(level: LogLevel, message: string, context?: any) {
    if (LEVEL_MAP[level] < LEVEL_MAP[this.currentLevel] || this.currentLevel === 'off') return;

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

  trace(message: string, context?: any) {
    if (process.env.NODE_ENV === 'development') {
      this.logToConsole('trace', message, context);
    }
    this.send('trace', message, context);
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
