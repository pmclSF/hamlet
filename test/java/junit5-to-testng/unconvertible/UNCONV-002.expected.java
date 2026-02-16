import org.testng.annotations.Test;

public class NestedTest {
    // HAMLET-TODO [UNCONVERTIBLE-NESTED]: JUnit 5 @Nested has no TestNG equivalent
    // Original: @Nested
    // Manual action required: Flatten nested test classes or use separate test classes
    @Nested
    class InnerTest {
        @Test
        public void testInner() {
            assert true;
        }
    }
}
