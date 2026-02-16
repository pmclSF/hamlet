import org.testng.annotations.AfterClass;
import org.testng.annotations.Test;

public class ClassTeardownTest {
    @AfterClass
    public static void tearDownClass() {
        System.out.println("class teardown");
    }

    @Test
    public void testSomething() {
        assert true;
    }
}
