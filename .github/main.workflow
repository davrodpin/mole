workflow "Mole Code Quality Checks" {
  on = "push"
  resolves = ["Coverage Report"]
}

action "Check Syntax" {
  uses = "./mole-checks"
  env = {
    GO_VERSION = "1.11.5"
  }
  args = [ "syntax" ]
}

action "Run Tests" {
  needs = [ "Check Syntax" ]
  uses = "./mole-checks"
  env = {
    GO_VERSION = "1.11.5"
  }
  args = [ "test" ]
}

action "Coverage Report" {
  needs = [ "Run Tests" ]
  uses = "./mole-checks"
  args = [ "report" ]
  secrets = ["GITHUB_TOKEN"]
}
