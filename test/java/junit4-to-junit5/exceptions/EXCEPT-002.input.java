import org.junit.Test;

public class TimeoutTest {
    @Test(timeout = 1000)
    public void testTimeout() {
        longRunningOperation();
    }
}
