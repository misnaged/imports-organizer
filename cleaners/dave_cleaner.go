package cleaners

import "os"

func (abs *AbstractSxTree) DaveCleaner(files []string) error {

	for i := range files {
		if ImportCheck(files[i]) {
			b, err := os.ReadFile(files[i])
			if err != nil {
				return fmt.Errorf("error ReadFile %w", err)
			}
			if err = abs.NewTree(b, files[i]); err != nil {
				return fmt.Errorf("NewTree in DaveCleaner %w", err)
			}

			f, err := decorator.Parse(b)
			if err != nil {
				return fmt.Errorf("decorator.Parse failed: %w", err)
			}

			for ii := range f.Imports {
				f.Imports[ii].Decorations().Start.Replace("")
				f.Imports[ii].Decorations().End.Replace("")
				f.Imports[ii].Decs.Start.Replace("")
				f.Imports[ii].Decs.End.Replace("")

				f.Imports[ii].Decs.NodeDecs.Start.Replace("")
				f.Imports[ii].Decs.NodeDecs.End.Replace("")
				f.Imports[ii].Decorations().Start.Replace("")
			}

			err = os.WriteFile(files[i], []byte(""), 0666)
			if err != nil {
				return fmt.Errorf(" WriteFile %w", err)
			}
			fileToOpen, err := os.OpenFile(files[i], os.O_APPEND|os.O_WRONLY, os.ModeAppend)
			if err != nil {
				return fmt.Errorf(" failed to open file: %w", err)
			}
			err = decorator.Fprint(fileToOpen, f)
			if err != nil {
				return fmt.Errorf(" printing decoration to the file failed: %w", err)
			}
		}
	}

	return nil
}
