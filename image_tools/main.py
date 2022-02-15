import argparse
import os
import re
import shutil
import textwrap
from urllib.parse import unquote


def get_args() :
    parser = argparse.ArgumentParser(
        description=textwrap.dedent(
            """
        A small tool for dealing with duplicate image attachments.
        It might be useful for your markdown's project and simplifying the attachment folder.
        """,
        ),
        formatter_class=argparse.RawTextHelpFormatter,
    )
    parser.add_argument(
        "--target-folder",
        type=str,
        default="assets",
        help="The name of the attachment folder.",
    )
    parser.add_argument(
        "--target-root", type=str, help="The root directory for storing your documents."
    )
    parser.add_argument(
        "--mode",
        type=str,
        default="list_useless",
        choices=["update_location", "list_useless", "delete_useless"],
        help=textwrap.dedent(
            """\
        - update_location: Move the image file to the image location indicated in the markdown document.
        - list_useless: List unreferenced image files.
        - delete_useless: Delete unreferenced image files.
        """
        ),
    )
    args = parser.parse_args()

    assert os.path.isabs(args.target_root)
    return args


def join_to_abs(*paths) :
    return os.path.abspath(os.path.join(*paths))


def get_paths_from_file_and_dir(target_root, target_folder) :
    # ![[assets/Pasted image 20220204120615.png]]
    md_image_pattern = re.compile(r"!\[.*?\]\((?!http)(.*?\.(jpg|png|jpeg|bmp|gif))\)",
                                  flags=re.IGNORECASE)
    wiki_image_pattern = re.compile(r"!\[\[(?!http)(.*?\.(jpg|png|jpeg|bmp|gif))\]\]",
                                    flags=re.IGNORECASE)

    paths_from_target_folder = []
    paths_from_dir_files = []
    for dir_path, dir_names, file_names in os.walk(target_root) :
        curr_dir_name = os.path.basename(dir_path)

        if curr_dir_name == target_folder :
            # Collect information from folders.
            paths_from_target_folder.extend(
                [join_to_abs(dir_path, n) for n in file_names]
            )
        else :
            # Collect information from files.
            for file_name in file_names :
                file_path = join_to_abs(dir_path, file_name)
                if not file_path.endswith(".md") :
                    continue

                with open(file_path, encoding="utf-8", mode="r") as f :
                    for line in f :
                        link = line.strip()
                        paths = md_image_pattern.findall(string=link)
                        paths += wiki_image_pattern.findall(string=link)
                        for image_path, image_ext in paths :
                            real_l = unquote(image_path)
                            paths_from_dir_files.append(join_to_abs(dir_path, real_l))

    paths_from_target_folder = list(set(paths_from_target_folder))
    paths_from_dir_files = list(set(paths_from_dir_files))
    return paths_from_target_folder, paths_from_dir_files


def main() :
    args = get_args()
    paths_from_target_folder, paths_from_dir_files = get_paths_from_file_and_dir(
        target_root=args.target_root, target_folder=args.target_folder
    )
    # target_folder中不存在于dir_files中的路径
    paths_difference_from_target_folder = set(paths_from_target_folder).difference(
        paths_from_dir_files
    )
    # dir_files中不存在于target_folder中的路径
    paths_difference_from_dir_files = set(paths_from_dir_files).difference(
        paths_from_target_folder
    )
    if args.mode == "update_location" :
        for path_from_file in paths_difference_from_dir_files :
            dir_path_from_file, file_name_from_file = os.path.split(path_from_file)
            for path_from_target_dir in paths_from_target_folder :
                dir_path_from_target_dir, file_name_from_target_dir = os.path.split(
                    path_from_target_dir)
                if file_name_from_file == file_name_from_target_dir :
                    if not os.path.exists(dir_path_from_file) :
                        os.makedirs(dir_path_from_file)
                    print(f"MOVE {path_from_target_dir} TO {path_from_file}...")
                    shutil.move(path_from_target_dir, dir_path_from_file)

                    if len(os.listdir(dir_path_from_target_dir)) == 0 :
                        print(f"DELETE EMPTY FOLDER {dir_path_from_target_dir}")
                        shutil.rmtree(dir_path_from_target_dir)
                    break
    elif args.mode == "list_useless" :
        print("unrefrenced image attachments:\n")
        print("\n".join(paths_difference_from_target_folder))
        print(f"number of image attachments in folder: {len(paths_from_target_folder)}")
        print(f"number of image attachments in md files: {len(paths_from_dir_files)}")
    elif args.mode == "delete_useless" :
        for path in paths_difference_from_target_folder :
            print(f"delete unrefrenced image attachment: {path}")
            # os.remove(path)
    else :
        raise NotImplementedError


if __name__ == "__main__" :
    main()
