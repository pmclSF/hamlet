import org.testng.annotations.AfterMethod;
import org.testng.annotations.Test;

public class AfterTest {
    @AfterMethod
    public void tearDown() {
        System.out.println("teardown");
    }

    @Test
    public void testSomething() {
        assert true;
    }
}
