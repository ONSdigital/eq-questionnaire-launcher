export default [
  {
    "env": {
      "browser": true,
      "es2021": true
    },
    "extends": "eslint:recommended",
    "parserOptions": {
      "ecmaVersion": "latest",
      "sourceType": "module"
    },
    "rules": {
      "no-unused-vars": "off",
      "no-undef": "off"
    }
  }
];