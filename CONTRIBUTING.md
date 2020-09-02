# Contributing

When contributing to this repository, please first discuss the change you wish
to make via [issues](https://github.com/davrodpin/mole/issues). 

Please note we have a [code of conduct](https://github.com/davrodpin/mole/blob/master/CODE_OF_CONDUCT.md),
please follow it in all your interactions with the project.

There are many different way you can contribute to this project:

* Implementing an [issue](https://github.com/davrodpin/mole/issues), whether
  that be a new feature or a bug that needs to be fixed
* Reporting a bug as an [issue](https://github.com/davrodpin/mole/issues)
* Discussing an idea
* Updating documentation
* Improving the site layout
* Implementing a test
* Testing the tool

Any contribution to this project is **highly** appreciated.

## Before opening a Pull Request

1. Make sure your change is covered by a test and all tests are passing (`make test`)
2. Remove linter warnings added by your code (`make lint`)
3. Validate your changes by doing manual tests using the [test enviornment](https://github.com/davrodpin/mole/blob/master/test-env/README.md)

## Pull Request Process

Once a Pull Request (PR) is created, [Github Actions](https://github.com/features/actions)
will automagically take care of validating your changes by building the project
and running the automated tests.
If either the build or the tests fail, the PR will not be merged and it is
expected that you fix the issues until the validation passes.

