import struct
import numpy as np
import tensorflow as tf
from random import shuffle
from PIL import Image
import matplotlib
import matplotlib.cm as cm
from scipy.special import softmax
import matplotlib.pyplot as plt
from tensorflow.keras import layers, initializers, regularizers, optimizers, callbacks

def load_samples(shuffle_samples):

    hash_size = 0
    with open('params.binary', "rb") as f:
        data = f.read()
        hash_size = int(data[0])**2
    
    X = []
    Y = []
    nb_samples = 0
    with open('X.binary', "rb") as f:
        data = f.read()
        nb_samples = len(data)//(hash_size*8)
        for i in range(nb_samples):
            tmp = [0.0 for i in range(hash_size)]
            buff = data[(i*hash_size)*8:(i+1)*hash_size*8]
            for j in range(hash_size):
                    tmp[j] = struct.unpack('d', buff[j*8:(j+1)*8])[0]
            X += [tmp]

    with open('Y.binary', "rb") as f:
        data = f.read()
        for i in data:
            tmp = [0, 0, 0, 0]
            tmp[int(i)] = 1
            Y += [tmp]

    if shuffle_samples:
        available = [i for i in range(nb_samples)]
        shuffle(available)
        X_shuffled = []
        Y_shuffled = []    
        for i in available:
            X_shuffled += [X[i]]
            Y_shuffled += [Y[i]]
        return np.array(X_shuffled), np.array(Y_shuffled)
    else:
        return np.array(X), np.array(Y)

def evaluate_model(k):

    X, Y = load_samples(True)

    nb_samples = len(X)

    features = len(X[0])

    split = nb_samples//k

    for i in range(k):
        print("k : {}/{}".format(i+1,k))

        x_train = np.array(X_suffled[:nb_samples-split*(i+1)] + X_suffled[nb_samples-split*i:])
        y_train = np.array(Y_suffled[:nb_samples-split*(i+1)] + Y_suffled[nb_samples-split*i:])
        x_val = np.array(X_suffled[nb_samples-split*(i+1):nb_samples-split*i])
        y_val = np.array(Y_suffled[nb_samples-split*(i+1):nb_samples-split*i])

        model = tf.keras.Sequential()
        model.add(tf.keras.layers.Dense(4, activation='softmax', kernel_initializer='he_normal', input_shape=(features,)))
        model.compile(
            optimizer='adam',
            loss='categorical_crossentropy',
            metrics='categorical_accuracy')
        history = model.fit(x_train, y_train, validation_data=(x_val, y_val), epochs=100, batch_size=32, verbose=0)

        for i in history.history.keys():
            print(i, history.history[i][-1])
        print()

def train_model():
    #TODO : 
    #Determine how resilient it is to shifting or holes
    #Look at what samples are not correctly classified (about 40 out of 8000), if there is a trend (like large gaps of 'N'),
    #Try without the Hilbert curve to see if it is actually needed
    #Remove irrelevant features to reduce the size of the model
    
    #Load the training data
    X, Y = load_samples(True)

    nb_samples = len(X)
    features = len(X[0])

    model = tf.keras.Sequential()
    model.add(tf.keras.layers.Dense(4, activation='softmax', kernel_initializer='he_normal', input_shape=(features,)))
    model.compile(
        optimizer='adam',
        loss='categorical_crossentropy',
        metrics='categorical_accuracy')

    history = model.fit(X, Y, epochs=100, batch_size=8, verbose=2, validation_split=0.0)

    predictions = model.predict(X, verbose=2)

    for i in range(len(model.layers)):
        layer = model.layers[i]
        #print(layer.get_config())
        weights = layer.get_weights()

        norm = matplotlib.colors.Normalize(vmin=np.min(weights[0]), vmax=np.max(weights[0]), clip=True)
        mapper = cm.ScalarMappable(norm=norm, cmap=cm.bwr)


        scaley = 80
        scalex = 8
        array = []
        for j in range(len(weights[0].T)):
            w = weights[0].T[j]
            for k in range(scaley):
                tmp0 = []
                for xx in w:
                    for u in range(scalex):
                        tmp0 += [[round(i*255) for i in mapper.to_rgba(xx)]]

                array += [tmp0]
        
        array = np.array(array, dtype=np.uint8)
        new_image = Image.fromarray(array)
        new_image.save('weights_layer_{}.png'.format(i))

        #Save both bias and weights in bianry and npy format
        data = []
        flattened_weights = weights[0].flatten()
        for j in range(len(flattened_weights)):
            data += struct.pack('d', flattened_weights[j])

        with open("weights_layer_{}".format(i), "wb") as f:
            f.write(bytearray(data))
            f.close()
        np.save('weights_layer_{}.npy'.format(i), weights[0])
        
        data = []
        flattened_bias = weights[1].flatten()
        for j in range(len(flattened_bias)):
            data += struct.pack('d', flattened_bias[j])

        with open("bias_layer_{}".format(i), "wb") as f:
            f.write(bytearray(data))
            f.close()
        np.save('bias_layer_{}.npy'.format(i), weights[1])

def final_model():
    ki = 'he_normal'
    bs = 32
    #regularizers
    l1=1e-6
    l2=1e-6
    #early stopping
    callback = callbacks.EarlyStopping(monitor='loss', patience=50)
    
    #Load the training data
    X, Y = load_samples(True)
    
    nb_samples = len(X)
    features = len(X[0])
    model = tf.keras.Sequential()
    model.add(tf.keras.layers.Dense(
        4,
        activation="softmax",
        kernel_initializer=ki,
        input_shape=(features,),
        kernel_regularizer=regularizers.l1_l2(l1=l1, l2=l2))
              )
    model.compile(
        optimizer="adam",
        loss='mean_squared_error',
        metrics='categorical_accuracy')
    
    print("results with featuers= ",features,", kernel_initializer = ", ki, ", batch_size = ", bs)
    
    history = model.fit(X, Y, epochs=256, batch_size=bs, verbose=2, validation_split=0.5, callbacks=[callback])

    SMALL_SIZE = 12
    MEDIUM_SIZE = 16
    BIGGER_SIZE = 20

    plt.rc('font', size=32)    
    plt.rc('axes', titlesize=32) 
    plt.rc('axes', labelsize=24)
    plt.rc('xtick', labelsize=16)
    plt.rc('ytick', labelsize=16)
    plt.rc('legend', fontsize=24) 
    plt.rc('figure', titlesize=16)
    plt.rcParams.update()

    # summarize history for accuracy
    plt.plot(history.history['categorical_accuracy'])
    plt.plot(history.history['val_categorical_accuracy'])
    plt.title('Model Accuracy')
    plt.ylabel('Accuracy')
    plt.xlabel('Epoch')
    plt.legend(['train', 'valid'], loc='lower right')
    plt.show()
    
    # summarize history for loss
    plt.plot(history.history['loss'])
    plt.plot(history.history['val_loss'])
    plt.title('Model Loss')
    plt.ylabel('Loss')
    plt.xlabel('Epoch')
    plt.legend(['train', 'valid'], loc='upper right')
    plt.show()
    
    #Save both bias and weights in bianry and npy format
    layer = model.layers[0]
    weights = layer.get_weights()
    data = []
    flattened_weights = weights[0].flatten()
    for j in range(len(flattened_weights)):
        data += struct.pack('d', flattened_weights[j])

    with open("weights_layer_{}".format(0), "wb") as f:
        f.write(bytearray(data))
        f.close()
    np.save('weights_layer_{}.npy'.format(0), weights[0])
    
    data = []
    flattened_bias = weights[1].flatten()
    for j in range(len(flattened_bias)):
        data += struct.pack('d', flattened_bias[j])

    with open("bias_layer_{}".format(0), "wb") as f:
        f.write(bytearray(data))
        f.close()
    np.save('bias_layer_{}.npy'.format(0), weights[1])
    
def test_model():
    from random import random
    
    X, Y = load_samples(False)

    bitsX = 5
    
    X *= (1<<bitsX)
    X = np.round(X)
    X /= (1<<bitsX)

    weights = np.load('weights_layer_0.npy')
    bias = np.load('bias_layer_0.npy')

    bitsW = 5
    
    weights *= (1<<bitsW)
    weights = np.round(weights)
    weights /= (1<<bitsW)

    bias *= (1<<bitsW)
    bias = np.round(bias)
    bias /= (1<<bitsW)

    print(weights[0])

    err = 0
    minDiff = 10
    minL = 0
    maxL = 0

    for i in range(len(X)):
        
        L = np.matmul(X[i], weights) + bias
        idx = L.tolist().index(np.max(L))

        L = L.tolist()
        L.sort()
        L= L[::-1]
    
        minL = min(minL, L[-1])
        maxL = max(maxL, L[0])
        minDiff = min(minDiff, L[0]-L[1])

        if idx != Y[i].tolist().index(1):
            err += 1

    print("errors :", err)
    print("minL", minL)
    print("maxL", maxL)
    print("minDiff", minDiff)

if __name__ == "__main__":
    #train_model()
    #evaluate_model(k=10)
    final_model()
    test_model()
