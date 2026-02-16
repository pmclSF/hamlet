import org.testng.annotations.Test;

public class DependsOnTest {
    @Test
    public void testLogin() {
        assert true;
    }

    @Test(dependsOnMethods = {"testLogin"})
    public void testDashboard() {
        assert true;
    }
}
