def test_cleanup():
    # HAMLET-TODO [UNCONVERTIBLE-ADDCLEANUP]: self.addCleanup has no direct pytest equivalent
    # Original: self.addCleanup(print, "cleanup")
    # Manual action required: Use a fixture with yield or request.addfinalizer
    self.addCleanup(print, "cleanup")
    assert True
