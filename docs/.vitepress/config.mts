import {defineConfig} from 'vitepress'

// https://vitepress.dev/reference/site-config
export default defineConfig({
    title: "Hygoal Docs & Hytale Protocol",
    description: "Documentation for the Hygoal Hytale server AND Hytale protocol definitions",
    themeConfig: {
        // https://vitepress.dev/reference/default-theme-config
        nav: [
            {text: 'Home', link: '/'},
            {text: "Hygoal Docs", link: '/hygoal/'},
            {text: "Hytale Protocol", link: '/protocol/'}
        ],

        sidebar: {
            '/hygoal/': [
                {
                    text: 'Hygoal Documentation',
                    items: [
                        {text: 'Introduction', link: '/hygoal/'},
                    ]
                }
            ],
            '/protocol/': [
                {
                    text: 'Hytale Protocol Definitions',
                    items: [
                        {text: 'Introduction', link: '/protocol/'},
                        {text: 'Handshake', link: '/protocol/handshake'},
                    ]
                },
            ]
        },

        socialLinks: [
            {icon: 'github', link: 'https://github.com/Floffah/hygoal'}
        ]
    }
})
