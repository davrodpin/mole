workflow "Mole Code Quality Checks" {
  on = "push"
  resolves = ["Check Syntax", "Publish Reports"]
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
  secrets = ["DROPBOX_TOKEN"]
}

action "Publish Reports" {
  needs = [ "Run Tests" ]
  uses = "./.github/actions/publish"
  secrets = ["DROPBOX_TOKEN"]
}
