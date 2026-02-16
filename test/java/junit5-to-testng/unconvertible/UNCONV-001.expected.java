import org.testng.annotations.Test;

public class DisplayNameTest {
    // HAMLET-TODO [UNCONVERTIBLE-DISPLAY-NAME]: JUnit 5 @DisplayName has no TestNG equivalent
    // Original: @DisplayName("A meaningful test name")
    // Manual action required: Use test method naming conventions or TestNG @Test(description = "...")
    @DisplayName("A meaningful test name")
    @Test
    public void testSomething() {
        assert true;
    }
}
