let s:dot_cmd             = "ninja --quiet -f vim.ninja"
let s:port_filename       = getpid().'.'.bufnr('%').'.'.'data.port'
let s:port                = substitute(readfile(s:port_filename)[0], ":", "", "")
let s:send_output_to_port = 'nc -tc 127.0.0.1 '.s:port

"silent execute 'w !dot -Tsvg|nc -tc 127.0.0.1 `cat '.getpid().'.'.bufnr('%').'.'.'data.port'.'|tr -d ":"`'
silent execute 'w !'.s:dot_cmd.'|'.s:send_output_to_port
