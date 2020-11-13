module.exports = {
  extends: [
    "airbnb-typescript/base",
    "prettier", // prettier should be last process
    "prettier/@typescript-eslint",
  ],
  plugins: ["@typescript-eslint", "import"],
  parser: "@typescript-eslint/parser",
  parserOptions: {
    project: ["./tsconfig.json"],
  },

  rules: {
    //
    // --- File layout ---
    //
    "max-classes-per-file": ["error", 3],
    "no-restricted-imports": [
      "error",
      {
        paths: [
          { name: "bluebird", message: "Use standard Promise" },
          { name: "lodash", message: "Use modern ES functions" },
          { name: "underscore", message: "Use modern ES functions" },
        ],
      },
    ],
    "sort-imports": "off", // Use eslint-plugin-import ("import/order") instead
    "import/order": [
      "error",
      {
        "newlines-between": "never",
        alphabetize: { order: "asc" },
      },
    ],
    "import/no-default-export": "error", // https://basarat.gitbook.io/typescript/main-1/defaultisbad , https://engineering.linecorp.com/ja/blog/you-dont-need-default-export/
    "import/prefer-default-export": "off", // Inverse of "import/no-default-export"

    //
    // --- Type / signature design ---
    //
    "class-methods-use-this": "off",
    "@typescript-eslint/no-use-before-define": "off",
    "@typescript-eslint/naming-convention": [
      "error",
      { selector: "default", format: ["camelCase"] },

      {
        selector: "variableLike",
        format: ["camelCase"],
        leadingUnderscore: "forbid",
      },
      {
        selector: "variable",
        format: [
          "camelCase",
          "UPPER_CASE", // const
          "PascalCase", // Variable of constructor function objects
        ],
        leadingUnderscore: "forbid",
      },
      {
        selector: "parameter",
        format: ["camelCase"],
        leadingUnderscore: "forbid",
      },

      { selector: "memberLike", format: ["camelCase"], leadingUnderscore: "allow" },
      {
        selector: "property",
        format: ["camelCase"],
        leadingUnderscore: "forbid",
      },

      { selector: "typeLike", format: ["PascalCase"] },
      { selector: "typeParameter", format: ["UPPER_CASE", "camelCase"] },

      { selector: "interface", format: ["PascalCase"] },
    ],

    //
    // --- Control flow ---
    //
    "no-continue": "off",

    //
    // --- Statements (within function) ---
    //
    "@typescript-eslint/no-unused-vars": [
      "error",
      { args: "none" }, // allow defining "abstract" members
    ],
    "no-restricted-syntax": [
      "error",
      // airbnb-base + permit for-of syntax
      { selector: "ForInStatement", message: "for..in loops iterate over the entire prototype chain, which is virtually never what you want. Use Object.{keys,values,entries}, and iterate over the resulting array." },
      // { selector: 'ForOfStatement', message: 'iterators/generators require regenerator-runtime, which is too heavyweight for this guide to allow them. Separately, loops should be avoided in favor of array iterations.' },
      { selector: "LabeledStatement", message: "Labels are a form of GOTO; using them makes code confusing and hard to maintain and understand." },
      { selector: "WithStatement", message: "`with` is disallowed in strict mode because it makes code impossible to predict and optimize." },
    ],
    "no-nested-ternary": "off",

    //
    // --- Standard API usage ---
    //
    "@typescript-eslint/no-floating-promises": "error", // Prevent missing await
    "@typescript-eslint/await-thenable": "error",
    "@typescript-eslint/promise-function-async": "error",
  },
};
