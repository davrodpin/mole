workflow "Mole Code Quality Checks" {
  on = "push"
  resolves = ["Check Code Quality"]
}

action "Check Code Quality" {
  uses = "./.github/actions/check"
  secrets = ["DROPBOX_TOKEN"]
}
