# Tools for managing my image attachments.

```shell
python .\main.py --help
usage: main.py [-h] [--target-folder TARGET_FOLDER] [--target-root TARGET_ROOT] [--mode {update_location,list_useless,delete_useless}]

A small tool for dealing with duplicate image attachments.
It might be useful for your markdown's project and simplifying the attachment folder.

optional arguments:
  -h, --help            show this help message and exit
  --target-folder TARGET_FOLDER
                        The name of the attachment folder.
  --target-root TARGET_ROOT
                        The root directory for storing your documents.
  --mode {update_location,list_useless,delete_useless}
                        - update_location: Move the image file to the image location indicated in the markdown document.
                        - list_useless: List unreferenced image files.
                        - delete_useless: Delete unreferenced image files.
```
