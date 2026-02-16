import org.testng.annotations.Test;

public class AssertThrowsTest {
    @Test(expectedExceptions = IllegalArgumentException.class)
    public void testException() {
        throwIllegalArg();
    }
}
