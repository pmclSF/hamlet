import org.testng.annotations.*;

public class GroupTest {
    @Test(groups = {"slow"})
    public void testSlow() {
        assert true;
    }
}
