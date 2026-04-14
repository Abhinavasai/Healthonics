// Karma configuration file, see link for more information
// https://karma-runner.github.io/1.0/config/configuration-file.html

// Use Puppeteer's bundled Chromium when CHROME_BIN is unset (CI, sandboxes without Chrome).
try {
  if (!process.env.CHROME_BIN && !process.env.CHROME_PATH) {
    process.env.CHROME_BIN = require('puppeteer').executablePath();
  }
} catch (_) {
  /* optional: install devDependency `puppeteer` for headless tests without system Chrome */
}

module.exports = function (config) {
  config.set({
    basePath: '',
    hostname: '127.0.0.1',
    frameworks: ['jasmine', '@angular-devkit/build-angular'],
    plugins: [
      require('karma-jasmine'),
      require('karma-chrome-launcher'),
      require('karma-jasmine-html-reporter'),
      require('karma-coverage'),
      require('@angular-devkit/build-angular/plugins/karma')
    ],
    client: {
      jasmine: {
        // you can add configuration options for Jasmine here
        // the possible options are listed at https://jasmine.github.io/api/edge/Configuration.html
        // for example, you can disable the random execution with `random: false`
        // or set a specific seed with `seed: 4321`
      },
    },
    jasmineHtmlReporter: {
      suppressAll: true // removes the duplicated traces
    },
    coverageReporter: {
      dir: require('path').join(__dirname, './coverage/frontend'),
      subdir: '.',
      reporters: [
        { type: 'html' },
        { type: 'text-summary' }
      ]
    },
    reporters: ['progress', 'kjhtml'],
    browsers: process.env.CI === 'true' ? ['ChromeHeadlessCI'] : ['Chrome'],
    customLaunchers: {
      // Use Chrome (not ChromeHeadless) + --headless=new: Puppeteer’s Chromium 109+ often fails with legacy --headless.
      ChromeHeadlessCI: {
        base: 'Chrome',
        flags: [
          '--headless=new',
          '--no-sandbox',
          '--disable-gpu',
          '--disable-dev-shm-usage',
          '--disable-software-rasterizer',
          '--mute-audio',
          '--remote-debugging-port=9333'
        ]
      }
    },
    restartOnFileChange: true,
    browserDisconnectTimeout: 20000,
    captureTimeout: 120000
  });
};
