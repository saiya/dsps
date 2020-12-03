module.exports = {
  preset: "ts-jest",
  testEnvironment: "node",

  silent: false,

  testMatch: ["<rootDir>/src/**/*.test.[jt]s"],
  coverageReporters: ["json", "html", "lcov", "text"],
};
