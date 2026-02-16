import org.junit.Test;

public class NestedExceptionTest {
    @Test(expected = MyService.NotFoundException.class)
    public void testNested() {
        service.findById(-1);
    }
}
