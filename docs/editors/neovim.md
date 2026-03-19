# Neovim

## nvim-lspconfig

### nvim < 0.11

```lua
require'lspconfig'.laravel_ls.setup{}
```

or custom config 

```lua
require'lspconfig'.laravel_ls.setup{
    -- Server-specific settings. See `:help lspconfig-setup`
    settings = {
        cmd = { …  },
    },
}
```

### nvim => 0.11

```lua
vim.lsp.config('laravel_ls', {
    cmd = { …  },
})
vim.lsp.enable('laravel_ls')
```

All settings can be found [here](https://github.com/neovim/nvim-lspconfig/blob/master/doc/configs.md#laravel_ls)

## native

The LSP server can be started like any other server via `vim.lsp.start` and an auto-command.

Just change the path to the correct directory on your filesystem

```lua
vim.api.nvim_create_autocmd("FileType", {
    pattern = { "php", "blade" },
    callback = function ()
        vim.lsp.start({
            name = "laravel-ls",

            -- if laravel ls is in your $PATH
            cmd = { 'laravel-ls' },

            -- Absolute path
            -- cmd = { '/path/to/laravel-ls/build/laravel-ls' },

            -- if you want to recompile everytime
            -- the language server is started.
            -- cmd = { '/path/to/laravel-ls/start.sh' },

            root_dir = vim.fn.getcwd(),
        })
    end
})
```

