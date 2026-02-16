import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.Assertions;
import java.util.List;
import java.util.ArrayList;

public class MixedImportTest {
    @Test
    public void testList() {
        List<String> items = new ArrayList<>();
        items.add("hello");
        Assertions.assertNotNull(items);
    }
}
