module.exports = (config) => {
  config.set({
    hostname: '127.0.0.1',
    listenAddress: '127.0.0.1',
    frameworks: ['jasmine'],
    customLaunchers: {
      ChromeHeadlessCI: {
        base: 'ChromeHeadless',
        flags: ['--no-sandbox', '--disable-dev-shm-usage', '--disable-gpu'],
      },
    },
  });
};
