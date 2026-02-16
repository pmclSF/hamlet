import org.testng.annotations.Test;

public class PriorityTest {
    @Test(priority = 1)
    public void testFirst() {
        assert true;
    }

    @Test(priority = 2)
    public void testSecond() {
        assert true;
    }
}
