workflow "Mole Code Quality Checks" {
  on = "push"
  resolves = ["Check Code Quality"]
}

action "Check Code Quality" {
  uses = "./.github/actions/check"
  env = {
    GO_VERSION = "1.11.5"
  }
  secrets = ["DROPBOX_TOKEN"]
}
