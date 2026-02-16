import org.testng.annotations.Listeners;
import org.testng.annotations.Test;

@Listeners(MyTestListener.class)
public class ListenersTest {
    @Test
    public void testSomething() {
        assert true;
    }
}
