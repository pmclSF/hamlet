import org.junit.jupiter.api.RepeatedTest;

public class RepeatedTestTest {
    // HAMLET-TODO [UNCONVERTIBLE-REPEATED-TEST]: JUnit 5 @RepeatedTest has no direct TestNG equivalent
    // Original: @RepeatedTest(5)
    // Manual action required: Use @Test(invocationCount = N) in TestNG
    @RepeatedTest(5)
    public void testRepeated() {
        assert true;
    }
}
