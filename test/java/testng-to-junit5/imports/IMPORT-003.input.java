import org.testng.annotations.Test;
import org.testng.Assert;
import java.util.List;
import java.util.ArrayList;

public class MixedImportTest {
    @Test
    public void testList() {
        List<String> items = new ArrayList<>();
        items.add("hello");
        Assert.assertNotNull(items);
    }
}
