import org.testng.annotations.Test;

public class ExpectedExceptionsTest {
    @Test(expectedExceptions = IllegalArgumentException.class)
    public void testException() {
        Integer.parseInt("not a number");
    }
}
