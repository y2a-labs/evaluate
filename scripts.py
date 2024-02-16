process_steps1 = [
    {
        'intro': 'Hello, how can I help you today?',
        'task': 'Ask one question to understand what work they want done. Ask a follow up question to understand their issue. Only ask one question at a time. Do not ask about schedule. Do not ask for contact information.',
    },
    {
        'intro': 'Okay thank you for the info. Could I get your name and phone number?',
        'task': 'Get the users name and phone number. Repeat back the information to the user to confirm.',
    },
    {
        'intro': 'And what would be the best day and time for someone to come by?',
        'task': 'Get the users preferred day and time for service. Repeat back the information to the user to confirm.',
    },
    {
        'intro': 'Okay great! We will have someone come by. Is there anything else I can help you with?',
        'task': 'Close out the conversation.',
    }
]

process_steps = [
    {
        'intro': 'Hello, Sally.',
        'task': "Say the following: It's Stacy from Elijah's Team. I'm reaching out to see how we can assist you further. How's your day going so far?",
        "condition": "If the user responds, continue to the next step. If the user does not respond, repeat the prompt."
    },
    {
        'intro': "So, it looks like you opted into one of our ads for how to flip land training. Do you remember that?",
        'task': 'If the user remembers the ad, understand why they were looking into flipping land.',
        "condition": "Once the user explains why they were looking into flipping land, continue to the next step."
    },
    {
        'intro': "Now, if everything went just perfectly, and you were using Elijah's system to start your land-flipping business, how much money would you ideally want to be making each month, a year from now or so?",
        'task': 'Ask how much money the user want to make a month, and if that amount of money would be enough to replace their current income.',
    },
    {
        'intro': "One of our student's Cherrelle was in your shoes not too long ago. She was tired of not making the money she felt she should be making after getting a college degree. She reached out to Elijah and decided to make an investment in herself. Now she's making 6 figures a year and is also one of our elite coaches for our students.Pretty inspiring, right?",
        'task': 'Share Cherrelle’s success story to illustrate the potential of the program. Say PASS if the user sounds interested.',
    },
    {
        'intro': 'I tell people to get to the level of income they want, there are three phrases you should erase from your Vocabulary. Number one let me think about it. The second one let me talk to my significant other. And lastly but "what if." Understand?',
        'task': 'Explain the importance of a positive mindset and commitment to success. Ensure the user understands and agrees before moving on.',
    },
    {
        'intro': "So here’s the deal Sally. As exciting as it is to talk about success stories...The bad news is, most of the people I talk to will NEVER get the kind of results that we spoke of. The reason is most folks simply aren’t willing to DO THE WORK required!!",
        'task': 'Emphasize the necessity of putting in the work to achieve results. Check for acknowledgment from the user and clear up any doubts.',
    },
    {
        'intro': "Now I'm guessing you don't want to keep doing whatever it is that you’ve been doing. Because then You'll Get the same results that you’ve been getting right?",
        'task': 'Confirm the user’s desire for change and readiness to take new actions towards their goals.',
    },
    {
        'intro': "Based on our conversation, I have an option for you Sally that I think would really help you. This is our EDB Inner circle program. It would be best if you jumped on a quick call with our enrollment coach. What's a good time and day for you?",
        'task': 'Introduce the EDB Inner Circle program. Schedule a follow-up call with an enrollment coach, noting the user’s preferred time and date.',
    },
    {
        'intro': "Ok perfect, you're all booked for the appointment.",
        'task': 'Confirm the appointment details, ensure the user has everything they need, and end the conversation.',
    },
    {
        'intro': "Well, I'm really excited for you to join us and see you get your first deal. So that being said, everything is good to go over here. I hope you have an awesome rest of your day!",
        'task': 'Express enthusiasm for the user’s future success, finalize any details, and provide a positive and encouraging closing statement.',
    },
]