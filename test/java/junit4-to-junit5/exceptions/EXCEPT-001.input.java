import org.junit.Test;

public class ExpectedExceptionTest {
    @Test(expected = IllegalArgumentException.class)
    public void testThrows() {
        Integer.parseInt("not a number");
    }
}
