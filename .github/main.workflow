workflow "Mole Code Quality Checks" {
  on = "push"
  resolves = ["Check Syntax", "Create Report"]
}

action "Check Syntax" {
  uses = "./.github/actions/syntax"
  env = {
    GO_VERSION = "1.11.5"
  }
}

action "Run Tests" {
  uses = "./.github/actions/test"
  env = {
    GO_VERSION = "1.11.5"
  }
}

action "Create Report" {
  needs = [ "Run Tests" ]
  uses = "./.github/actions/report"
  env = {
    GO_VERSION = "1.11.5"
  }
  secrets = ["GITHUB_TOKEN"]
}
