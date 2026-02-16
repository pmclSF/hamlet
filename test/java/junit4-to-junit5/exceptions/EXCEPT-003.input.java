import org.junit.Test;

public class CombinedTest {
    @Test(expected = RuntimeException.class)
    public void testExpected() {
        throwRuntime();
    }

    @Test(timeout = 500)
    public void testTimed() {
        quickOperation();
    }
}
