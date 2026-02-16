import org.junit.jupiter.api.Test;
import static org.junit.jupiter.api.Assertions.assertThrows;

public class NestedExceptionTest {
    @Test
    public void testNested() {
        assertThrows(MyService.NotFoundException.class, () -> {
            service.findById(-1);
        });
    }
}
