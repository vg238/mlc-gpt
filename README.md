# mlc-gpt
MLC-GPT is a simple RAG (Retrieval Augmented Generation) app created using LlamaIndex and GPT-4.
It can be used to answer questions about the Major League Cricket (MLC) tournament held in the
United States during summer 2023. It uses LlamaIndex's AutoVectorQueryEngine to dynamically decide
the best combination of relational and vector databases to use for answering a given question.

**Try it**: https://www.mlcguru.com

Steps to run your own version of the app:

```
>> export OPENAI_API_KEY=<Your OPENAI API Key>
>> pip install -r requirements.txt
>> python app.py
```

**High-level overview of LlamaIndex's AutoVectorQueryEngine**

1. Initialized with two query tools. A SQL query tool and a Vector query tool.
2. Sends the input question and the descriptions of the query tools to a LLM and lets it choose the first tool to apply.
3. If the first tool is SQL, then, sends the input question and the database table schema information to a LLM and asks it to generate a SQL query.
4. Queries the database using the SQL query.
5. Sends the SQL response and request information to a LLM and asks it to generate a natural language response.
6. Sends the natural language response and the original question to a LLM and asks for a follow-up question to the vector tool (if needed)
7. If the LLM has a follow up question to the vector tool, then, generates the embeddings for the question and retrives top 2 documents from the vector database
8. Sends this context and the follow up question to a LLM and asks it to generate a natural language response.
9. Sends the original question, the natural language response from SQL tool, the natural language response from vector tool to a LLM and asks it to generate a final answer.
10. Return the final answer to the user.

## Data Preparation

The app uses contents of the data and chorma_db folders. This data is already prepared and ready for the app to consume. It was collected from the Major League Cricket's website and Google News.

For more details about the data collection and preparation process, see the code in the data_prep folder.
