declare module "pino/browser" {
  import type { Logger, LoggerOptions } from "pino";

  type BrowserPinoFactory = (options?: LoggerOptions) => Logger;

  const pino: BrowserPinoFactory;
  export default pino;
}
