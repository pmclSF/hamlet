import org.junit.Test;
import org.junit.Rule;
import org.junit.rules.TemporaryFolder;

public class TempFolderRuleTest {
    @Rule
    public TemporaryFolder folder = new TemporaryFolder();

    @Test
    public void testTempFile() throws Exception {
        folder.newFile("test.txt");
    }
}
