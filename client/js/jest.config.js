module.exports = {
  preset: "ts-jest",
  testEnvironment: "node",
  testMatch: ["<rootDir>/src/**/*.test.[jt]s"],
  coverageReporters: ["json", "html", "lcov", "text"],
};
