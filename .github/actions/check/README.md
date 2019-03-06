# Mole Quality Checker

Github Action to report if a given change meets the project quality standards.

This action generates the following reports:

  * Files that have formatting issues
  * Code Coverage for the target commit
  * Code Coverage difference between the target commit and its predecessor to
    inform if the coverage has increased or decreased

## Environment Variables

N/A

## Secrets

  * `DROPBOX_TOKEN`: Authentication token allow reports to be uploaded to
    Dropbox.

## Required Arguments

N/A

## Optional Arguments

N/A
