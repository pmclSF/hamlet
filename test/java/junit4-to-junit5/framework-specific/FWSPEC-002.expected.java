import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.AfterEach;
import org.junit.jupiter.api.Assertions;
import org.junit.jupiter.api.Disabled;
import java.util.List;
import java.util.ArrayList;

public class FullConversionTest {
    private List<String> items;

    @BeforeEach
    public void setUp() {
        items = new ArrayList<>();
    }

    @AfterEach
    public void tearDown() {
        items.clear();
    }

    @Test
    public void testAdd() {
        items.add("hello");
        Assertions.assertEquals(1, items.size());
        Assertions.assertTrue(items.contains("hello"));
    }

    @Disabled
    @Test
    public void testNotReady() {
        Assertions.fail("not implemented");
    }
}
