import org.testng.annotations.AfterClass;
import org.testng.annotations.Test;

public class AfterAllTest {
    @AfterClass
    public static void tearDownClass() {
    }

    @Test
    public void testSomething() {
        assert true;
    }
}
