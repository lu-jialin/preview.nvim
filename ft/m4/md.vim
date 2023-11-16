let s:m4_cmd             = "m4 -I$HOME/.config -D__root__=\"`git rev-parse --show-toplevel`\" -D__git__=\"`git rev-parse --show-toplevel|xargs basename`\" -D__relative_root__=\"`realpath -m --relative-to=$PWD $(git rev-parse --show-toplevel)`\" -D__path__=\"`echo $(realpath .)|grep -oP \"(?<=$(git rev-parse --show-toplevel)/).*\"`\"|pandoc -f markdown -s --toc --preserve-tabs --standalone -H $HOME/.lulix/dvl/css/github.css"
let s:port_filename       = getpid().'.'.bufnr('%').'.'.'data.port'
let s:port                = substitute(readfile(s:port_filename)[0], ":", "", "")
let s:send_output_to_port = 'nc -tc 127.0.0.1 '.s:port

"silent execute 'w !dot -Tsvg|nc -tc 127.0.0.1 `cat '.getpid().'.'.bufnr('%').'.'.'data.port'.'|tr -d ":"`'
execute 'w !'.s:m4_cmd.'|'.s:send_output_to_port
