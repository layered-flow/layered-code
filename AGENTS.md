* Namespace tools correclty, eg. lc_js_lint should be in the 'internal/tools/js' directory
* Stick to the current paradigm of adding a tool making it callable from both the CLI and the MCP (see list_apps.go for reference)
* If required, use existing helpers in 'internal/config/config.go' like EnsureAppsDirectory and ensure traversal to other folders is protected - do not duplicate the logic in 'config.go'
* If required, use existing helpers in 'internal/helpers/validation.go' like ValidateAppName - do not duplicate the logic in 'validation.go'
* MCP handlers must return errors (never panic) so LLMs get useful error messages instead of MCP faults/alert boxes
* Do not create any tests unless explicity asked to, I want to check the functionality first before creating tests
* If I do ask for a test, Keep it lean and only test what is pertinent to that tool. Do not re-test library or internal Go functionality as that will already be covered in tests for Go or that library
* Do not add anything to the readme unless explicity asked to
* Run all existing tests and ensure they are passing when you are done.
* Rebuild the project using the command `make` when all tests are passing